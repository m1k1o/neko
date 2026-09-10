const path = require('path')

module.exports = {
  productionSourceMap: false,
  css: {
    loaderOptions: {
      sass: {
        additionalData: `
          @import "@/assets/styles/_variables.scss";
        `,
      },
    },
  },
  publicPath: './',
  assetsDir: './',
  configureWebpack: {
    // The emoji sprite and font fallbacks are intentional static assets. Keep
    // performance checks enabled, but size them against the current asset
    // budget so normal builds do not report known baseline warnings.
    performance: {
      maxAssetSize: 5 * 1024 * 1024,
      maxEntrypointSize: 2 * 1024 * 1024,
    },
    resolve: {
      alias: {
        vue$: 'vue/dist/vue.esm.js',
        '~': path.resolve(__dirname, 'src/'),
      },
    },
  },
  devServer: {
    allowedHosts: 'all',
  },
}
