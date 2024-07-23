import { rpc } from "./rpc.js";

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

export function AddMarker(req: AddMarkerReq): Promise<Empty> {
    return rpc("api.Api", "AddMarker", req);
}

export function DeleteRange(req: DeleteRangeReq): Promise<Empty> {
    return rpc("api.Api", "DeleteRange", req);
}

export function UpdateVariables(req: UpdateVariablesReq): Promise<Empty> {
    return rpc("api.Api", "UpdateVariables", req);
}

export function ReadVariables(req: ReadVariablesReq): Promise<ReadVariablesResp> {
    return rpc("api.Api", "ReadVariables", req);
}

export function LoadMarkers(req: LoadMarkersReq): Promise<LoadMarkersResp> {
    return rpc("api.Api", "LoadMarkers", req);
}
