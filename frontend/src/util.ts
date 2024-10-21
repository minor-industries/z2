import {Notyf} from 'notyf';

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

export function localDate(): string {
    const now = new Date();
    const year = now.getFullYear();
    const month = (now.getMonth() + 1).toString().padStart(2, '0'); // Months are zero-indexed, so +1
    const day = now.getDate().toString().padStart(2, '0');

    return `${year}-${month}-${day}`;
}

const notyf = new Notyf({
    duration: 5000,
    ripple: false,
    dismissible: true,
});

export function notify(type: 'success' | 'error', message: string) {
    switch (type) {
        case "success":
            notyf.success(message);
            return;
        case "error":
            notyf.error(message);
            return;
        default:
            notyf.error("unknown alert type");
            return;
    }
}
