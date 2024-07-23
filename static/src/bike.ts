import {DrawCallbackArgs, Graph, synchronize} from 'rtgraph';
import * as api from "./api";
import {dygraphs} from 'dygraphs'
import {v4 as uuidv4} from 'uuid';

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

type MarkerType = "b" | "e";

export type Marker = {
    id: string,
    type: MarkerType
    ref: string
    timestamp: number
}

// TODO: showModal should be redone
export function setupBikeAnalysis(date: string | null) {
    const second = 1000;

    const markers: Marker[] = [];

    const saveMarkers = async () => {
        for (let i = 0; i < markers.length; i++) {
            const m = markers[i];
            await api.AddMarker({
                marker: m,
            });
        }
    };

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
        drawCallback: (args: DrawCallbackArgs) => {
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

            const markerType = prompt("(b)egin or (e)nd?")
            switch (markerType) {
                case 'b':
                case 'e':
                    break;
                case null:
                    return;
                default:
                    alert("unknown markerType");
                    return;
            }

            if (point.xval === undefined) {
                console.log("no xval");
                return;
            }

            markers.push({
                id: uuidv4(),
                type: markerType,
                ref: "bike",
                timestamp: point.xval,
            })

            const annotations = markers.map(m => ({
                series: 'y1',
                x: m.timestamp,
                shortText: m.type,
                text: m.type, // TODO:
                attachAtBottom: true,
                dblClickHandler: function (
                    annotation: dygraphs.Annotation,
                    point: dygraphs.Point,
                    dygraph: any,
                    event: MouseEvent,
                ) {
                    console.log(annotation);
                },
            }));

            (g1.dygraph as any).setAnnotations(annotations);
            (g2.dygraph as any).setAnnotations(annotations);

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
                    return api.DeleteRange({
                        start: start,
                        end: end,
                    });
                }
                return;
            case "KeyS":
                if (confirm("save markers?")) {
                    saveMarkers();
                }
                return;
            default:
                return Promise.resolve();
        }
    };

    document.addEventListener("keydown", ev => {
        keyDown(ev);
    })
}