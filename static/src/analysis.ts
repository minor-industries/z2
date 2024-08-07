import {DrawCallbackArgs, Graph, synchronize} from 'rtgraph';
import * as api from "./api";
import {dygraphs} from 'dygraphs'
import {v4 as uuidv4} from 'uuid';
import * as agg from "./aggregate"


export type Marker = {
    id: string
    type: string
    ref: string
    timestamp: number
}

export async function saveMarkers(markers: Marker[]) {
    for (let i = 0; i < markers.length; i++) {
        const m = markers[i];
        await api.AddMarker({
            marker: m,
        });
    }
}

export type AnalysisArgs = {
    date: string
    ref: string
    seriesNames: string[],
    title: string,
    ylabel: string,
};

export function setupAnalysis(args: AnalysisArgs) {
    const second = 1000;

    const markers: Marker[] = [];

    const seriesOpts = {
        y2: {strokeWidth: 1.0},
        y3: {strokeWidth: 1.0, color: "red"},
        y4: {strokeWidth: 1.0, color: "red"},
        y5: {strokeWidth: 1.0, color: "red"},
    }

    const g1 = new Graph(document.getElementById("graphdiv0")!, {
        seriesNames: args.seriesNames,
        title: "Avg Speed",
        ylabel: "speed (km/h)",
        windowSize: null,
        height: 250,
        maxGapMs: 5 * second,
        series: seriesOpts,
        disableScroll: true,
        date: args.date,
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
        date: args.date,
        drawCallback: (args: DrawCallbackArgs) => {
            console.log(
                "max Value", agg.maxV(args, 0),
                "delta t", agg.deltaT(args, 0),
                "avg HR", agg.avgV(args, 1),
            );
        }
    });

    function updateAnnotations() {
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
    }

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
                ref: args.ref,
                timestamp: point.xval,
            });

            updateAnnotations();
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
                    saveMarkers(markers);
                }
                return;
            default:
                return Promise.resolve();
        }
    };

    api.LoadMarkers({date: args.date, ref: args.ref}).then(resp => {
        markers.push(...resp.markers);
        updateAnnotations();
    })

    document.addEventListener("keydown", ev => {
        keyDown(ev);
    })
}



