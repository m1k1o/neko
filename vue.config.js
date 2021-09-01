const path = require('path')

module.exports = {
  productionSourceMap: false,
  configureWebpack: {
    resolve: {
      alias: {
        vue$: 'vue/dist/vue.esm.js',
        '~': path.resolve(__dirname, 'src/'),
      },
    },
  },
  devServer: {
    disableHostCheck: true,
    proxy: {
      '^/api': {
        target: 'http://' + process.env.NEKO_HOST + ':' + process.env.NEKO_PORT + '/',
      },
    },
  },
}
