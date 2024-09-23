export function streamEvents(
    path: string,
    callback: (data: string) => void
): void {
    const es = new EventSource(path);
    es.addEventListener("play-sound", (event) => {
        callback(event.data);
    });
}
