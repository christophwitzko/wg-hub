const isDev = process.env.NODE_ENV !== "production";

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "export",
  env: {
    VERSION: process.env.VERSION || "0.0.0-dev",
  },
  ...(isDev
    ? {
        rewrites() {
          return [
            {
              source: "/api/:path*",
              destination: "http://192.168.0.254/api/:path*",
            },
          ];
        },
      }
    : {}),
};

export default nextConfig;
