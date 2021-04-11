
export type PowerSummary = {
    meter: number;
    year: number;
    consumption: number;
    month?: number;
    day?: number;
    hour?: number;
}

export type PowerDetail = {
    meter: number;
    year: number;
    month: number;
    day: number;
    seconds: number;
    consumption: number;
}