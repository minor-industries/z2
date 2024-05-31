interface TwirpError {
    code: string
    msg: string
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