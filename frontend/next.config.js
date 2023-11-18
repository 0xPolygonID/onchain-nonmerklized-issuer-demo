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
}

module.exports = nextConfig
