interface TwirpError {
    code: string
    msg: string
}

export async function rpc(
    service: string,
    method: string,
    req: any,
) {
    const response = await fetch(`/twirp/${service}/${method}`, {
        method: "POST",
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(req),
    });

    if (!response.ok) {
        const body: TwirpError = await response.json();
        const msg = `rpc error: http code=${response.status}; code=${body.code}; msg=${body.msg};`;
        throw new Error(msg);
    }

    return response.json();
}
