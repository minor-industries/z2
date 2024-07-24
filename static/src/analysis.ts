import {DrawCallbackArgs} from 'rtgraph';
import * as api from "./api";


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


