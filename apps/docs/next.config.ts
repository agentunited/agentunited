import type { NextConfig } from "next";
import nextra from "nextra";

const withNextra = nextra({});

const nextConfig: NextConfig = {
  output: "standalone",
  async redirects() {
    return [
      {
        source: "/docs/python-sdk",
        destination: "/docs/sdks/python",
        permanent: true,
      },
      {
        source: "/docs/typescript-sdk",
        destination: "/docs/sdks/typescript",
        permanent: true,
      },
    ];
  },
};

export default withNextra(nextConfig);
