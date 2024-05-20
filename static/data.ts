interface TwirpError {
    code: string
    msg: string
}

export interface GetDatesReq {

}

export interface GetDatesResp {
    dates: {
        datestr: string
        count: number
    }[]
}

export type Empty = {};

export type DeleteRangeReq = {
    start: bigint
    end: bigint
}

export function GetDates(req: GetDatesReq): Promise<GetDatesResp> {
    return rpc("api.Calendar", "GetDates", req)
}

export function DeleteRange(req: DeleteRangeReq): Promise<Empty> {
    return rpc("api.Api", "DeleteRange", req)
}

export function rpc(
    service: string,
    method: string,
    req: any,
) {
    return fetch(`/twirp/${service}/${method}`, {
        method: "POST",
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(req),
    }).then(response => {
        if (!response.ok) {
            return response.json().then((body: TwirpError) => {
                const msg = `rpc error: http code=${response.status}; code=${body.code}; msg=${body.msg};`;
                return Promise.reject(Error(msg));
            })
        }
        return response.json();
    })
}