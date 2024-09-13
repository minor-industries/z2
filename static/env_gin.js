import {calendar, DefaultApiClient, runOnce, streamEvents} from "/dist/z2-bundle.js";

async function maybeStartFrontendBLE() {
}

async function setup() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient(),
        maybeStartFrontendBLE,
        streamEvents,
    };
}

export const getEnv = runOnce(setup);