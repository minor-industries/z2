export function runOnce<T extends (...args: any[]) => Promise<any>>(asyncFn: T): T {
    let initialized = false;
    let resultPromise: ReturnType<T>;

    return (async function (...args: Parameters<T>) {
        if (!initialized) {
            initialized = true;
            resultPromise = asyncFn(...args) as ReturnType<T>; // Explicitly cast the result to ReturnType<T>
        }
        return resultPromise;
    }) as T;
}
