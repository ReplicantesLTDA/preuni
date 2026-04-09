// Proxy /v1/* to the NGINX API gateway (localhost:8080) during local development.
// This lets the frontend call relative paths like /v1/auth/register without CORS issues.
if (config.devServer) {
    config.devServer.proxy = [
        {
            context: ['/v1'],
            target: 'http://localhost:8080',
            changeOrigin: true,
        }
    ];
}
