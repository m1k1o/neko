const path = require('path')
const webpack = require('webpack')

module.exports = {
  productionSourceMap: false,
  publicPath: './',
  assetsDir: './',  
  configureWebpack: {
    resolve: {
      alias: {
        vue$: 'vue/dist/vue.esm.js',
        '~': path.resolve(__dirname, 'src/'),
      },
    },
    plugins: [
      new webpack.NormalModuleReplacementPlugin(
        /(.*)__KEYBOARD__/,
        function(resource){
          resource.request = resource.request
            .replace(/__KEYBOARD__/, process.env.KEYBOARD || 'guacamole');
        },
      ),
    ],
  },
  devServer: {
    allowedHosts: "all",
    proxy: {
      '^/api': {
        target: 'http://' + process.env.NEKO_HOST + ':' + process.env.NEKO_PORT + '/',
      },
    },
  },
}
