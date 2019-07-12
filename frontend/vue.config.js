module.exports = {
  devServer: {
    proxy: {
      '^/api/v1': {
        target: 'http://localhost:8080',
        changeOrigin: false
      }
    }
  }
}
