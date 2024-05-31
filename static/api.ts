import {rpc} from "./rpc.js";

export interface GetDatesReq {
}

export interface DateInfo {
    datestr: string;
    count: number;
}

export interface GetDatesResp {
    dates: DateInfo[];
}

export type Empty = {};

export interface DeleteRangeReq {
    start: bigint;
    end: bigint;
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

export function GetDates(req: GetDatesReq): Promise<GetDatesResp> {
    return rpc("api.Calendar", "GetDates", req);
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

