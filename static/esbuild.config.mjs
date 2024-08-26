import {build} from 'esbuild';

build({
    entryPoints: ['./dist/index.js'],
    bundle: true,
    outfile: './dist/z2-bundle.js',
    format: 'esm',
    minify: false,
    plugins: [],
}).catch(() => process.exit(1));
