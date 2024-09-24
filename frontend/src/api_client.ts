import {rpc} from "./rpc.js";

export type Marker = {
    id: string;
    type: string;
    ref: string;
    timestamp: number;
};

export interface AddMarkerReq {
    marker: Marker;
}

export interface DeleteRangeReq {
    start: number;
    end: number;
}

export interface Variable {
    name: string;
    value: number;
    present: boolean;
}

export interface UpdateVariablesReq {
    variables: Variable[];
}

export interface ReadVariablesReq {
    variables: string[];
}

export interface ReadVariablesResp {
    variables: Variable[];
}

export interface LoadMarkersReq {
    ref: string;
    date: string; // Date in the format "2024-01-02"
}

export interface LoadMarkersResp {
    markers: Marker[];
}

export type Empty = {};


export interface ApiClient {
    addMarker(req: AddMarkerReq): Promise<Empty>;

    deleteRange(req: DeleteRangeReq): Promise<Empty>;

    updateVariables(req: UpdateVariablesReq): Promise<Empty>;

    readVariables(req: ReadVariablesReq): Promise<ReadVariablesResp>;

    loadMarkers(req: LoadMarkersReq): Promise<LoadMarkersResp>;

    disconnectBluetoothDevices(req: Empty): Promise<Empty>;
}

export class DefaultApiClient implements ApiClient {
    public async addMarker(req: AddMarkerReq): Promise<Empty> {
        return rpc("api.Api", "AddMarker", req);
    }

    public async deleteRange(req: DeleteRangeReq): Promise<Empty> {
        return rpc("api.Api", "DeleteRange", req);
    }

    public async updateVariables(req: UpdateVariablesReq): Promise<Empty> {
        return rpc("api.Api", "UpdateVariables", req);
    }

    public async readVariables(req: ReadVariablesReq): Promise<ReadVariablesResp> {
        return rpc("api.Api", "ReadVariables", req);
    }

    public async loadMarkers(req: LoadMarkersReq): Promise<LoadMarkersResp> {
        return rpc("api.Api", "LoadMarkers", req);
    }

    public async disconnectBluetoothDevices(req: Empty): Promise<Empty> {
        return rpc("api.Api", "DisconnectBluetoothDevices", req);
    }
}
