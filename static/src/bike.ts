import {DrawCallbackArgs, Graph, synchronize} from 'rtgraph';
import {DeleteRange} from "./api";
import {dygraphs} from 'dygraphs'

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

function maxV(args: DrawCallbackArgs, i: number) {
    let max = Number.MIN_VALUE;

    const {i0, i1, V} = select(args, i);

    for (let i = i0; i < i1; i++) {
        if (V[i] > max) {
            max = V[i];
        }
    }

    return max === Number.MIN_VALUE ? NaN : max;
}

function avgV(args: DrawCallbackArgs, i: number) {
    const {i0, i1, V} = select(args, i);

    let count = 0;
    let sum = 0;

    for (let i = i0; i < i1; i++) {
        count++;
        sum += V[i];
    }

    return count === 0 ? NaN : sum / count;
}

function deltaT(args: DrawCallbackArgs, i: number) {
    const {lo, hi} = select(args, i);
    return (hi - lo) / 1000.0 / 60.0
}

// TODO: showModal should be redone
export function setupBikeAnalysis(date: string | null) {
    const second = 1000;

    const seriesOpts = {
        y2: {strokeWidth: 1.0},
        y3: {strokeWidth: 1.0, color: "red"},
        y4: {strokeWidth: 1.0, color: "red"},
        y5: {strokeWidth: 1.0, color: "red"},
    }

    const g1 = new Graph(document.getElementById("graphdiv0")!, {
        seriesNames: [
            "bike_avg_speed_long | time-bin",
            "bike_avg_speed_short | time-bin",
            "bike_instant_speed_min | time-bin",
            "bike_instant_speed_max | time-bin",
            "bike_target_speed | time-bin",
        ],
        title: "Avg Speed",
        ylabel: "speed (km/h)",
        windowSize: null,
        height: 250,
        maxGapMs: 5 * second,
        series: seriesOpts,
        disableScroll: true,
        date: date,
    });

    const g2 = new Graph(document.getElementById("graphdiv1")!, {
        seriesNames: [
            "heartrate | avg 2m triangle | time-bin",
            "heartrate | time-bin"
        ],
        title: "Heartrate",
        ylabel: "bpm",
        windowSize: null,
        height: 250,
        series: seriesOpts,
        maxGapMs: 5 * second,
        disableScroll: true,
        date: date,
        drawCallback: (args) => {
            console.log(
                "max Value", maxV(args, 0),
                "delta t", deltaT(args, 0),
                "avg HR", avgV(args, 1),
            );
        }
    });

    const graphs = [
        g1.dygraph,
        g2.dygraph,
    ];

    (g1.dygraph as any).updateOptions({
        pointClickCallback: function (e: MouseEvent, point: dygraphs.Point) {

            const kind = prompt("(b)egin or (e)nd?")
            switch (kind) {
                case 'b':
                case 'e':
                    break;
                case null:
                    return;
                default:
                    alert("unknown kind");
                    return;
            }

            const an1: dygraphs.Annotation = {
                series: 'y1',
                x: point.xval,
                shortText: kind,
                text: 'Marker',
                attachAtBottom: true,
                dblClickHandler: function (
                    annotation: dygraphs.Annotation,
                    point: dygraphs.Point,
                    dygraph: any,
                    event: MouseEvent,
                ) {
                    console.log(annotation);
                },
            };
            console.log(an1);
            (g1.dygraph as any).setAnnotations([an1]);
            (g2.dygraph as any).setAnnotations([an1]);

            console.log(point);
        },
    })

    const sync = synchronize(graphs, {
        selection: true,
        zoom: true,
        range: false,
    });

    const keyDown = (ev: KeyboardEvent) => {
        switch (ev.code) {
            case "KeyD":
                const range = (graphs[0] as any).xAxisRange();
                const dateRange = [new Date(range[0]), new Date(range[1])];
                console.log(range);

                const start = Math.floor(range[0]);
                const end = Math.ceil(range[1]);
                console.log(start, end);

                const prompt = `are you sure you want to delete the currently visible range? \n${dateRange[0]}\n${dateRange[1]}`;
                const ok = confirm(prompt)
                if (ok) {
                    alert(`deleting ${dateRange}`)
                    return DeleteRange({
                        start: start,
                        end: end,
                    });
                }
                return;
            case "KeyK":
            default:
                return Promise.resolve();
        }
    };

    document.addEventListener("keydown", ev => {
        keyDown(ev);
    })
}