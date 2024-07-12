import {build} from 'esbuild';

build({
    entryPoints: ['./dist/index.js'],
    bundle: true,
    outfile: './dist/bundle.js',
    format: 'esm',
    minify: true,
    plugins: [],
}).catch(() => process.exit(1));
