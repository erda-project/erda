const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const {CleanWebpackPlugin} = require('clean-webpack-plugin');

module.exports = {
    entry: {
        index: ['./src/index.tsx']
    },
    output: {
        filename: '[name].bundle.js',
        path: path.resolve(__dirname, 'public')
    },
    resolve: {
        extensions: [".ts", ".tsx", ".js", ".json"]
    },
    // externals: {
    //     "react": "React",
    //     "react-dom": "ReactDOM"
    // },
    module: {
        rules: [
        {
            test: /\.tsx?$/, 
            loader: "awesome-typescript-loader",
            exclude: /node_modules/
        },
        { 
            test: /\.js$/, 
            enforce: "pre", 
            loader: "source-map-loader" 
        },
        {
            test: /\.css$/,
            use: ['style-loader', 'css-loader']
        },
        {
            test: /\.scss$/,
            use: ['style-loader', 'css-loader', 'sass-loader']
        },
        {
            test: /\.less$/,
            use: [
                'style-loader', 'css-loader',
                {
                    loader: 'less-loader',
                    options: {
                        javascriptEnabled: true
                    }
                }
            ]
        },
        {
            test: /\.(png|svg|jpg|gif|ico)$/,
            use: ['file-loader']
        },
        {
            test: /\.(woff|woff2|eot|ttf|otf)$/,
            use: ['file-loader']
        }]
    },
    devtool: 'source-map',
    devServer: {
        contentBase: './public',
        hot: true
    },
    plugins: [
        new CleanWebpackPlugin(),
        new HtmlWebpackPlugin({
            template: 'src/index.html',
            favicon: 'src/assets/favicon.ico'
        })
    ]
};