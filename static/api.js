import { rpc } from "./rpc.js";
export function DeleteRange(req) {
    return rpc("api.Api", "DeleteRange", req);
}
export function UpdateVariables(req) {
    return rpc("api.Api", "UpdateVariables", req);
}
export function ReadVariables(req) {
    return rpc("api.Api", "ReadVariables", req);
}
