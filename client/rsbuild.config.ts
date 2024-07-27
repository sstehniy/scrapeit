import { defineConfig } from "@rsbuild/core";
import { pluginReact } from "@rsbuild/plugin-react";

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
        target: "http://golang-scraper:3457",
        changeOrigin: true,
        pathRewrite: (path) => path.replace(/^\/api/, ""),
        secure: false,
      },
    },
  },
});
