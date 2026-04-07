// Run the webpack dev server on port 3000 so it doesn't conflict with the
// NGINX API gateway which binds 0.0.0.0:8080.
if (config.devServer) {
    config.devServer.port = 3000;
}
