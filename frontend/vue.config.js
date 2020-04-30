module.exports = {
  devServer: {
    port: 8081,
    proxy: {
      '^/api/v1': {
        target: 'http://localhost:8080',
        changeOrigin: false
      }
    }
  }
}
