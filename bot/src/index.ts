import { Hono } from "hono";
import { logger } from "hono/logger";
import { prettyJSON } from "hono/pretty-json";
import { trimTrailingSlash } from "hono/trailing-slash";

import { createClient } from "redis";

import { Telegraf } from "telegraf";

if (!process.env.REDIS_URL) {
	throw new Error("REDIS_URL is required");
}

const redis = createClient({
	url: process.env.REDIS_URL,
});

redis.on("error", (err) => {
	console.error(err);
	process.exit(1);
});
await redis.connect();

if (!process.env.TELEGRAM_BOT_TOKEN) {
	throw new Error("TELEGRAM_BOT_TOKEN is required");
}

const bot = new Telegraf(process.env.TELEGRAM_BOT_TOKEN);

bot.command("start", async (ctx) => {
	console.log("start command received");
	if (!ctx.from.id) {
		return ctx.reply("Error: User ID not found");
	}
	const found = await redis.get(`user:${ctx.from.id}`);
	if (found === "active") {
		return ctx.reply("You are already active");
	}

	await redis.set(`user:${ctx.from.id}`, "active");

	return ctx.reply("You are now active");
});

bot.on("chat_member", async (ctx) => {
	console.log("chat_member event received", ctx.chatMember);
});

bot.launch();
const app = new Hono();

app.use(prettyJSON());
app.use(logger());
app.use(trimTrailingSlash());

app.get("/", (c) => {
	return c.text("Hello Hono!");
});

type SearchResult = {
	uniqueHash: string;
	endpointName: string;
	url: string;
	fields: { fieldName: string; value: string }[];
	status: "new" | "updated";
	imageUrl?: string; // Add this field to include an optional image URL
};

type SearchFilter = {
	fieldName: string;
	operator: string;
	value: string;
};

type RequestBody = {
	results: SearchResult[];
	filters: SearchFilter[];
	groupName: string;
};

const MAX_MESSAGES_PER_SECOND = 20;
let messageCount = 0;
let lastResetTime = Date.now();

async function sendMessageWithDebounce(
	bot: Telegraf,
	userId: string,
	method: string,
	...args: any[]
) {
	const now = Date.now();
	if (now - lastResetTime >= 750) {
		// Reset the counter every second
		messageCount = 0;
		lastResetTime = now;
	}

	if (messageCount >= MAX_MESSAGES_PER_SECOND) {
		// If the limit is reached, wait until the next second
		await new Promise((resolve) =>
			setTimeout(resolve, 2000 - (now - lastResetTime)),
		);
		messageCount = 0;
		lastResetTime = Date.now();
	}

	messageCount++;
	console.log({ args });
	// Call the appropriate method on the bot instance
	return (bot.telegram as any)[method](userId, ...args);
}

async function sendResultsAsMediaGroup(
	userId: string,
	chunk: { endpointName: string; status: string; results: SearchResult[] },
	mainMessage: string,
) {
	const resultsWithImages = chunk.results.filter((result) => result.imageUrl);
	const textOnlyResults = chunk.results.filter((result) => !result.imageUrl);

	// Split results with images into groups of 10
	const mediaGroups = splitIntoMediaGroups(resultsWithImages);

	// Send each media group as individual photos with a message
	for (const group of mediaGroups) {
		for (const result of group) {
			const formattedResult = `-------------\nResults for endpoint: ${chunk.endpointName}\n${
				mainMessage
			}\n${formatSingleResult(result)}\n-------------`;

			try {
				// Attempt to send the photo with the caption
				await sendMessageWithDebounce(
					bot,
					userId,
					"sendPhoto",
					result.imageUrl!,
					{
						caption: formattedResult, // Escape HTML special characters
						parse_mode: "HTML",
					},
				);
			} catch (error) {
				console.log({ "Error sendPhoto Message": error });

				// If sending the photo fails, send the result as a text message
				try {
					await sendMessageWithDebounce(
						bot,
						userId,
						"sendMessage",
						formattedResult, // Escape HTML special characters
						{
							parse_mode: "HTML",
							disable_web_page_preview: true,
						},
					);
				} catch (error) {
					console.log({ "Error sendMessage Message": error });
				}
			}
		}
	}

	// Send text-only results
	if (textOnlyResults.length > 0) {
		const textMessage = `-------------\nResults for endpoint: ${chunk.endpointName}\n${
			mainMessage
		}\n<b>${chunk.status} results:</b>\n${textOnlyResults.map(formatSingleResult).join("")}\n-------------`;

		try {
			await sendMessageWithDebounce(bot, userId, "sendMessage", textMessage, {
				parse_mode: "HTML",
				disable_web_page_preview: true,
			});
		} catch (error) {
			console.log({ error });
		}
	}
}

app.post("/send-notification", async (c) => {
	const body = await c.req.json<RequestBody>();

	const activeUsers = await redis.keys("user:*");

	const formattedResultChunks = formatResults(body);

	for (const user of activeUsers) {
		const userId = user.split(":")[1];
		// Send main message with group name and filters
		const mainMessage = formatMainMessage(body.groupName, body.filters);

		for (const chunk of formattedResultChunks) {
			if (chunk.results.length > 0) {
				await sendResultsAsMediaGroup(userId, chunk, mainMessage);
			}
		}
	}

	return c.json({ success: true });
});

const formatMainMessage = (groupName: string, filters: SearchFilter[]) => {
	const filterText = filters.map((filter) => {
		return `${escapeText(filter.fieldName)} ${escapeText(filter.operator)} ${escapeText(filter.value)}`;
	});
	return `GROUP: ${escapeText(groupName)}\nFilters: ${filterText.join(", ")}\n`;
};

function splitIntoMediaGroups(results: SearchResult[]): SearchResult[][] {
	const MAX_MEDIA_GROUP_SIZE = 10;
	const groups: SearchResult[][] = [];

	for (let i = 0; i < results.length; i += MAX_MEDIA_GROUP_SIZE) {
		groups.push(results.slice(i, i + MAX_MEDIA_GROUP_SIZE));
	}

	return groups;
}

function formatResults(body: RequestBody) {
	const groupedResults = body.results.reduce(
		(acc, result) => {
			acc[result.endpointName] = acc[result.endpointName] || [];
			acc[result.endpointName].push(result);
			return acc;
		},
		{} as Record<string, SearchResult[]>,
	);

	const messages: Array<{
		endpointName: string;
		status: string;
		results: SearchResult[];
	}> = [];

	for (const [endpointName, results] of Object.entries(groupedResults)) {
		const newResults = results.filter((r) => r.status === "new");
		const updatedResults = results.filter((r) => r.status === "updated");

		if (newResults.length > 0)
			messages.push({ endpointName, status: "New", results: newResults });
		if (updatedResults.length > 0)
			messages.push({
				endpointName,
				status: "Updated",
				results: updatedResults,
			});
	}

	return messages;
}

function formatSingleResult(result: SearchResult): string {
	const MAX_FIELD_VALUE_LENGTH = 100;
	const cutValue = (value: string) => {
		if (value.length > MAX_FIELD_VALUE_LENGTH) {
			return value.slice(0, MAX_FIELD_VALUE_LENGTH) + "...";
		}
		return value;
	};
	const fields = result.fields
		.map(
			(field) =>
				`<b>${escapeText(field.fieldName)}</b>: ${escapeText(cutValue(field.value))}`,
		)
		.join("\n");
	return `<b><a href="${result.url}">URL</a></b>\n${fields}\n`;
}

function escapeText(text: unknown): string {
	if (typeof text !== "string") {
		console.log({ text });
		return text as string;
	}
	return text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#039;");
}

export default {
	port: 5005,
	fetch: app.fetch,
};
