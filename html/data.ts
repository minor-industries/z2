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

export function GetDates(req: GetDatesReq): Promise<GetDatesResp> {
    return rpc("GetDates", req)
}

export function rpc(
    method: string,
    req: any,
) {
    return fetch(`/twirp/api.Calendar/${method}`, {
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