import {calendar, DefaultApiClient, runOnce} from "/dist/z2-bundle.js";

async function setup() {
    return {
        apiClient: new DefaultApiClient(),
        calendarClient: new calendar.DefaultApiClient()
    };
}

export const getEnv = runOnce(setup);