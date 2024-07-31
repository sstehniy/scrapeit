import { defineConfig } from "@rsbuild/core";
import { pluginReact } from "@rsbuild/plugin-react";

// @ts-ignore
console.log("BACKEND_URL", process.env.BACKEND_URL);

export default defineConfig({
	html: {
		template: "./index.html",
	},
	source: {
		entry: {
			index: "./src/main.tsx",
		},
	},
	plugins: [
		pluginReact({
			enableProfiler: true,
		}),
		,
	],
	performance: {
		printFileSize: true,
		buildCache: false,
	},
	server: {
		port: 3456,
		host: "0.0.0.0",
		proxy: {
			"/api": {
				// @ts-ignore
				target: process.env.BACKEND_URL,
				changeOrigin: true,
				pathRewrite: (path) => path.replace(/^\/api/, ""),
				secure: false,
			},
		},
	},
});
