import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ command, mode }) => {
  // Load env file based on `mode` in the current working directory.
  const env = loadEnv(mode, process.cwd(), "");
  
  const isProduction = mode === "production";
  
  return {
    plugins: [react()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
      },
    },
    server: {
      port: 3000,
      proxy: {
        "/api": {
          target: "http://localhost:8080",
          changeOrigin: true,
        },
        "/ws": {
          target: "ws://localhost:8080",
          ws: true,
        },
      },
    },
    build: {
      outDir: "dist",
      sourcemap: isProduction ? false : true,
      minify: isProduction ? "terser" : false,
      rollupOptions: {
        output: {
          manualChunks: {
            vendor: ["react", "react-dom"],
            router: ["react-router-dom"],
            supabase: ["@supabase/supabase-js"],
            websocket: ["socket.io-client"],
          },
          // Optimize chunk naming for production
          chunkFileNames: isProduction 
            ? "assets/js/[name]-[hash].js"
            : "assets/js/[name].js",
          entryFileNames: isProduction 
            ? "assets/js/[name]-[hash].js"
            : "assets/js/[name].js",
          assetFileNames: isProduction 
            ? "assets/[ext]/[name]-[hash].[ext]"
            : "assets/[ext]/[name].[ext]",
        },
      },
      chunkSizeWarningLimit: 1000,
      // Production optimizations
      ...(isProduction && {
        target: "es2015",
        cssCodeSplit: true,
        reportCompressedSize: false,
      }),
    },
    preview: {
      port: 3000,
      host: true,
    },
    test: {
      globals: true,
      environment: "jsdom",
      setupFiles: ["./src/test/setup.ts"],
    },
    // Environment variable handling
    define: {
      __APP_VERSION__: JSON.stringify(env.VITE_APP_VERSION || "1.0.0"),
      __APP_NAME__: JSON.stringify(env.VITE_APP_NAME || "Collaborative Bucket List"),
      __ENVIRONMENT__: JSON.stringify(env.VITE_ENVIRONMENT || "development"),
    },
  };
});
