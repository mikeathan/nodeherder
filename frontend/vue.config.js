const { defineConfig } = require('@vue/cli-service');

module.exports = defineConfig({
  devServer: {
    port: 4100,
  },
  transpileDependencies: true,
  configureWebpack: {
    devtool: 'source-map',
  },
});
