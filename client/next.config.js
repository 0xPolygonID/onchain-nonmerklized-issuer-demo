const nextConfig = {
  async headers() {
    return [
      {
        source: "/",
        headers: [
          {
            key: "Access-Control-Allow-Origin",
            value: "*",
          },
        ],
      },
    ];
  },
  publicRuntimeConfig: {
    NEXT_PUBLIC_ISSUER_URL: process.env.NEXT_PUBLIC_ISSUER_URL,
  },

  webpack: (config, options) => {
    const opt = {
      ...config,
      optimization: {
        ...config.optimization,
        splitChunks: false,
        // runtimeChunk: false,
        // minimize: false,
        // emitOnErrors: false,
        // usedExports: false
      },
    };
    // console.log(opt);
    return opt;
  },
};

module.exports = nextConfig;
