import {calendar, DefaultApiClient, runOnce} from "/dist/z2-bundle.js";

async function setup() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient(),
        maybeStartFrontendBLE() {
            return Promise.resolve();
        }
    };
}

export const getEnv = runOnce(setup);