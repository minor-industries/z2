import {build} from 'esbuild';
import fs from 'fs';
import path from 'path';

const copyFiles = () => {
    const filesToCopy = `
        bumper-control.css
        cal.css
        calendar.html
        dygraph.css
        hrm.html
        index.html
        rtgraph.css
        sse.html
    `.trim().split(/\s+/)
    const destDir = './z2/';

    filesToCopy.forEach(file => {
        const destPath = path.join(destDir, path.basename(file));
        fs.copyFileSync(file, destPath);
    });
};

build({
    entryPoints: ['./build/index.js'],
    bundle: true,
    outfile: './z2/z2-bundle.js',
    format: 'esm',
    minify: false,
    plugins: [],
})
    // .then(copyFiles);
