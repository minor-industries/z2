import {rpc} from "./rpc.js";

import {
    AddMarkerReq,
    DeleteRangeReq,
    Empty,
    LoadMarkersReq,
    LoadMarkersResp,
    ReadVariablesReq,
    ReadVariablesResp,
    UpdateVariablesReq
} from "./api";

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
