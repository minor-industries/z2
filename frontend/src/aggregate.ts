import {DrawCallbackArgs} from 'rtgraph';

// TODO: DrawCallbackArgs is needlessly coupled here

function select(args: DrawCallbackArgs, i: number) {
    return {
        lo: args.lo,
        hi: args.hi,
        i0: args.indices[i][0],
        i1: args.indices[i][1],
        Ts: args.series[i].Timestamps,
        V: args.series[i].Values
    }
}

export function maxV(args: DrawCallbackArgs, i: number) {
    let max = Number.MIN_VALUE;

    const {i0, i1, V} = select(args, i);

    for (let i = i0; i < i1; i++) {
        if (V[i] > max) {
            max = V[i];
        }
    }

    return max === Number.MIN_VALUE ? NaN : max;
}

export function avgV(args: DrawCallbackArgs, i: number) {
    const {i0, i1, V} = select(args, i);

    let count = 0;
    let sum = 0;

    for (let i = i0; i < i1; i++) {
        count++;
        sum += V[i];
    }

    return count === 0 ? NaN : sum / count;
}

export function deltaT(args: DrawCallbackArgs, i: number) {
    const {lo, hi} = select(args, i);
    return (hi - lo) / 1000.0 / 60.0
}
