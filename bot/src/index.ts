import { Hono } from "hono";
import { logger } from "hono/logger";
import { prettyJSON } from "hono/pretty-json";
import { trimTrailingSlash } from "hono/trailing-slash";

import { createClient } from "redis";

import { Telegraf } from "telegraf";
import type { InputMediaPhoto } from "telegraf/types";

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

app.post("/send-notification", async (c) => {
	const body = await c.req.json<RequestBody>();
	await Bun.write("./test.json", JSON.stringify(body, null, 2));

	const activeUsers = await redis.keys("user:*");
	console.log(activeUsers);
	const formattedResultChunks = formatResults(body);
	const mainMessage = formatMainMessage(body.groupName, body.filters);

	for (const user of activeUsers) {
		const userId = user.split(":")[1];
		await bot.telegram.sendMessage(userId, ">>>>>>>>>>>>>>>>");
		await bot.telegram.sendMessage(userId, mainMessage);
		for (const chunk of formattedResultChunks) {
			if (chunk.results.length > 0) {
				await sendResultsAsMediaGroup(userId, chunk);
			}
		}
		await bot.telegram.sendMessage(userId, "<<<<<<<<<<<<<<");
	}

	return c.json({ success: true });
});

const formatMainMessage = (groupName: string, filters: SearchFilter[]) => {
	const filterText = filters.map((filter) => {
		return `${filter.fieldName} ${filter.operator} ${filter.value}`;
	});
	return `New results in GROUP ${groupName} with filters: ${filterText.join(", ")}`;
};

const MAX_MESSAGE_LENGTH = 4096;

async function sendResultsAsMediaGroup(
	userId: string,
	chunk: { endpointName: string; status: string; results: SearchResult[] },
) {
	const resultsWithImages = chunk.results.filter((result) => result.imageUrl);
	const textOnlyResults = chunk.results.filter((result) => !result.imageUrl);

	// Split results with images into groups of 10
	const mediaGroups = splitIntoMediaGroups(resultsWithImages);

	// Send each media group
	for (const group of mediaGroups) {
		const mediaGroupPayload = group.map((result) => ({
			type: "photo" as const,
			// biome-ignore lint/style/noNonNullAssertion: <explanation>
			media: result.imageUrl!,
			caption: formatSingleResult(result),
			parse_mode: "HTML" as const,
		}));

		await bot.telegram.sendMediaGroup(userId, mediaGroupPayload);
	}

	// Send text-only results
	if (textOnlyResults.length > 0) {
		const textMessage = `<b>Results for endpoint: ${chunk.endpointName}</b>\n\n<b>${chunk.status} results:</b>\n\n${textOnlyResults.map(formatSingleResult).join("")}`;
		await bot.telegram.sendMessage(userId, textMessage, {
			parse_mode: "HTML",
			link_preview_options: {
				is_disabled: true,
			},
		});
	}
}

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
	const fields = result.fields
		.map((field) => `<b>${field.fieldName}</b>: ${field.value}`)
		.join("\n");
	return `<b><a href="${result.url}">URL</a></b>\n${fields}\n\n`;
}

export default {
	port: 5005,
	fetch: app.fetch,
};
