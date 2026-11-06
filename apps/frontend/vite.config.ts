import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3010,
    strictPort: true,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@taskify/openapi": path.resolve(__dirname, "../../packages/openapi/src"),
      "@taskify/zod": path.resolve(__dirname, "../../packages/zod/src"),
    },
  },
});
