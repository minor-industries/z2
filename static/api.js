import { rpc } from "./rpc.js";
export function GetDates(req) {
    return rpc("api.Calendar", "GetDates", req);
}
export function DeleteRange(req) {
    return rpc("api.Api", "DeleteRange", req);
}
export function UpdateVariables(req) {
    return rpc("api.Api", "UpdateVariables", req);
}
export function ReadVariables(req) {
    return rpc("api.Api", "ReadVariables", req);
}
