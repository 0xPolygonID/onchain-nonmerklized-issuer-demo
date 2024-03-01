const nextConfig = {
    async headers() {
        return [
          {
            source: '/',
            headers: [
              {
                key: 'Access-Control-Allow-Origin',
                value: '*',
              }
            ],
          },
        ]
      },
    publicRuntimeConfig: {
      NEXT_PUBLIC_ISSUER_URL: process.env.NEXT_PUBLIC_ISSUER_URL,
    },  
}

module.exports = nextConfig
