export function streamEvents(
    path: string,
    callbacks: { [key: string]: (data: string) => void }
): void {
    const es = new EventSource(path);

    for (const eventType in callbacks) {
        if (callbacks.hasOwnProperty(eventType)) {
            es.addEventListener(eventType, (event: MessageEvent) => {
                callbacks[eventType](event.data);
            });
        }
    }
}
