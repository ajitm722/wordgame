const path = require("path");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ForkTsCheckerWebpackPlugin = require("fork-ts-checker-webpack-plugin");

const frontendDir = path.resolve(__dirname, "frontend");

module.exports = (env, argv) => {
  const isProd = argv.mode === "production";

  return {
    entry: path.join(frontendDir, "index.tsx"),
    output: {
      path: path.resolve(__dirname, "assets"),
      filename: isProd ? "bundle.[contenthash:8].js" : "bundle.js",
      publicPath: "/assets/",
      clean: true,
    },
    resolve: {
      extensions: [".ts", ".tsx", ".js", ".jsx"],
      modules: [frontendDir, "node_modules"],
    },
    module: {
      rules: [
        {
          test: /\.tsx?$/,
          use: "ts-loader",
          exclude: /node_modules/,
        },
        {
          test: /\.scss$/,
          use: [
            isProd ? MiniCssExtractPlugin.loader : "style-loader",
            "css-loader",
            "postcss-loader",
            "sass-loader",
          ],
        },
        {
          test: /\.css$/,
          use: [
            isProd ? MiniCssExtractPlugin.loader : "style-loader",
            "css-loader",
          ],
        },
      ],
    },
    plugins: [
      new HtmlWebpackPlugin({
        template: path.join(frontendDir, "templates", "react.ejs"),
        filename: path.resolve(__dirname, "assets", "index.html"),
        inject: true,
      }),
      ...(isProd
        ? [
            new MiniCssExtractPlugin({
              filename: "bundle.[contenthash:8].css",
            }),
          ]
        : []),
      new ForkTsCheckerWebpackPlugin({
        typescript: {
          configFile: path.resolve(__dirname, "tsconfig.json"),
        },
      }),
    ],
    devtool: isProd ? false : "eval-source-map",
    performance: {
      maxAssetSize: 400000,
      maxEntrypointSize: 400000,
    },
  };
};
