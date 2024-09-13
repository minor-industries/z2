export function streamEvents(
    path: string,
    callback: (data: string) => void
): void {
    const es = new EventSource(path);
    es.onmessage = (event) => {
        callback(event.data);
    };
}
