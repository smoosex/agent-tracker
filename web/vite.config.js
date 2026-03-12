import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const base = process.env.AGENT_TRACKER_BASE_PATH || "/";

export default defineConfig({
  base,
  plugins: [react()],
  server: {
    port: 20001,
    proxy: {
      "/api": {
        target: "http://localhost:10001",
        changeOrigin: true,
      },
      "/rss": {
        target: "http://localhost:10001",
        changeOrigin: true,
      },
    },
  },
});
