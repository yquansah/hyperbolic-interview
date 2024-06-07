import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    proxy: {
      "/argo": "http://localhost:8080",
    },
    origin: "http://localhost:5173",
  },
});
