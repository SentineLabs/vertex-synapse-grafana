const path = require('path');
const webpack = require('webpack');
const CopyWebpackPlugin = require('copy-webpack-plugin');
const ESLintPlugin = require('eslint-webpack-plugin');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');
const ReplaceInFileWebpackPlugin = require('replace-in-file-webpack-plugin');

const packageJson = require('./package.json');

const DIST_DIR = path.resolve(__dirname, './dist');
const SRC_DIR = path.resolve(__dirname, './src');

const common = {
  context: __dirname,
  devtool: 'source-map',
  externals: [
    'lodash',
    'jquery',
    'moment',
    'slate',
    'prismjs',
    '@grafana/slate-react',
    '@grafana/data',
    '@grafana/ui',
    '@grafana/runtime',
    function ({ request }, callback) {
      const prefix = 'grafana/';
      if (request && request.indexOf(prefix) === 0) {
        return callback(null, request.substr(prefix.length));
      }
      callback();
    },
  ],
  plugins: [
    new ESLintPlugin({
      extensions: ['.ts', '.tsx'],
      lintDirtyModulesOnly: true, // don't lint on start, only lint changed files
    }),
    new ForkTsCheckerWebpackPlugin({
      async: true, // don't block webpack emit
      typescript: {
        mode: 'write-references',
        memoryLimit: 4096,
      },
    }),
    new CopyWebpackPlugin({
      patterns: [
        { from: 'plugin.json', to: '.' },
        { from: 'README.md', to: '.' },
        { from: 'CHANGELOG.md', to: '.', noErrorOnMissing: true },
        { from: 'LICENSE', to: '.', noErrorOnMissing: true },
        { from: 'src/img', to: 'img', noErrorOnMissing: true },
        { from: 'src/dashboards', to: 'dashboards', noErrorOnMissing: true },
      ],
    }),
    new webpack.EnvironmentPlugin({
      NODE_ENV: 'development',
    }),
    new ReplaceInFileWebpackPlugin([
      {
        dir: DIST_DIR,
        files: ['plugin.json'],
        rules: [
          {
            search: '%VERSION%',
            replace: packageJson.version,
          },
          {
            search: '%TODAY%',
            replace: new Date().toISOString().substring(0, 10),
          },
        ],
      },
    ]),
  ],
  resolve: {
    extensions: ['.js', '.jsx', '.ts', '.tsx'],
    modules: [SRC_DIR, 'node_modules'],
    unsafeCache: true,
  },
  stats: {
    children: false,
    warningsFilter: /export .* was not found in/,
  },
  target: ['web', 'es5'],
};

module.exports = (env = {}) => [
  {
    ...common,
    entry: {
      module: path.resolve(SRC_DIR, 'module.ts'),
    },
    mode: env.production ? 'production' : 'development',
    module: {
      // Note: order is bottom-to-top and/or right-to-left
      rules: [
        {
          test: /\.tsx?$/,
          use: {
            loader: 'swc-loader',
            options: {
              jsc: {
                parser: {
                  syntax: 'typescript',
                  tsx: true,
                  decorators: false,
                  dynamicImport: true,
                },
              },
            },
          },
          exclude: /node_modules/,
        },
        {
          test: /\.(sa|sc|c)ss$/,
          use: ['style-loader', 'css-loader', 'sass-loader'],
        },
      ],
    },
    output: {
      clean: true,
      filename: '[name].js',
      library: {
        type: 'amd',
      },
      path: DIST_DIR,
    },
  },
];
