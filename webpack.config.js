const path = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');


module.exports = {
    entry: './templates/src/app.js',
    output: {
        path: path.resolve(__dirname, 'templates/dist'),
        filename: 'bundle.js'
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: [MiniCssExtractPlugin.loader, 'css-loader'],
            }
        ]
    },
    plugins: [new MiniCssExtractPlugin({ filename: 'style.css' })],
}