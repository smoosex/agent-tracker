import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
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
