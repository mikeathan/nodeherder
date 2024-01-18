const { defineConfig } = require("@vue/cli-service");

module.exports = defineConfig({
  devServer: {
    port: 4100,
  },
  transpileDependencies: true,
  configureWebpack: {
    devtool: "source-map",
  },
  // configureWebpack: {
  //   resolve: {
  //     extensions: ['.ts', '.tsx', '.vue', '.js', '.json'],
  //   },
  //   module: {
  //     rules: [
  //       {
  //         test: /\.tsx?$/,
  //         loader: 'ts-loader',
  //         options: { appendTsSuffixTo: [/\.vue$/] },
  //         exclude: /node_modules/,
  //       },
  //     ],
  //   },
  //},
});
