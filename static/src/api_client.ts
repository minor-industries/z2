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
    AddMarker(req: AddMarkerReq): Promise<Empty>;

    DeleteRange(req: DeleteRangeReq): Promise<Empty>;

    UpdateVariables(req: UpdateVariablesReq): Promise<Empty>;

    ReadVariables(req: ReadVariablesReq): Promise<ReadVariablesResp>;

    LoadMarkers(req: LoadMarkersReq): Promise<LoadMarkersResp>;
}

export class DefaultApiClient implements ApiClient {
    public async AddMarker(req: AddMarkerReq): Promise<Empty> {
        return rpc("api.Api", "AddMarker", req);
    }

    public async DeleteRange(req: DeleteRangeReq): Promise<Empty> {
        return rpc("api.Api", "DeleteRange", req);
    }

    public async UpdateVariables(req: UpdateVariablesReq): Promise<Empty> {
        return rpc("api.Api", "UpdateVariables", req);
    }

    public async ReadVariables(req: ReadVariablesReq): Promise<ReadVariablesResp> {
        return rpc("api.Api", "ReadVariables", req);
    }

    public async LoadMarkers(req: LoadMarkersReq): Promise<LoadMarkersResp> {
        return rpc("api.Api", "LoadMarkers", req);
    }
}
