<script lang="ts">
    import {
        Chart,
        LinearScale,
        BarController,
        LineController,
        CategoryScale,
        BarElement,
        Title,
        Tooltip,
        PointElement,
        LineElement,
    } from "chart.js";
    import { onDestroy, onMount } from "svelte";
    import type { PowerSummary } from "./powersummary";

    Chart.register(
        LinearScale,
        BarController,
        LineController,
        CategoryScale,
        BarElement,
        Title,
        Tooltip,
        PointElement,
        LineElement
    );

    // refreshRateInMS sets the refresh rate of the live view in milliseconds
    export let refreshRateInMS: number = 60_000;

    async function fetchPower() {
        // fetch Power entries
        const today: Date = new Date();
        const queries: Promise<Response>[] = [];

        // fetch 3 days worth of data to ensure we are able to get 48 entries
        for (let i = 0; i < 3; i++) {
            const y: number = today.getFullYear();
            const m: number = today.getMonth() + 1;
            const d: number = today.getDate();

            // @ts-ignore: SERVICE_URL is included in rollup build.
            const url = SERVICE_URL + `/${y}/${m}/${d}`;
            queries.push(fetch(url));
            today.setDate(d - 1);
        }

        const resp = await Promise.all<Response, Response, Response>([
            queries[0],
            queries[1],
            queries[2],
        ]);

        let usage = await Promise.all<
            PowerSummary[],
            PowerSummary[],
            PowerSummary[]
        >([resp[0].json(), resp[1].json(), resp[2].json()]);

        const allUsage = [...usage[2], ...usage[1], ...usage[0]];

        const now: Date = new Date();

        let entries: number[] = new Array(48);
        let labels: string[] = new Array(48);
        let j = allUsage.length - 1;
        for (let i = 47; i >= 0; i--) {
            const h = now.getHours();
            labels[i] = `${h}h`;

            entries[i] = 0;
            if (j >= 0 && allUsage[j].hour == h) {
                entries[i] = allUsage[j].consumption;
                j--;
            }

            now.setHours(h - 1);
        }

        let avg = new Array<number>(3);
        chart.data.labels = labels;
        chart.data.datasets[0].data = entries;
        chart.data.datasets[1].data = entries.map<number>((v, i) => {
            avg[i % 3] = v;
            return avg.reduce<number>((tot, v) => tot + v, 0) / 3;
        });
        chart.options.scales.y.max = Math.max(...entries) + 1;
        chart.update();
    }

    let ctx: HTMLCanvasElement;
    let timer: number;
    let chart: Chart;

    onMount(() => {
        fetchPower();
        timer = setInterval(() => fetchPower(), refreshRateInMS);

        chart = new Chart(ctx, {
            type: "bar",
            data: {
                datasets: [
                    {
                        type: "bar",
                        label: "kWh",
                        backgroundColor: "rgba(0, 128, 0, 0.6)",
                        data: null,
                        borderWidth: 1,
                    },
                    {
                        type: "line",
                        label: "3-hour moving average",
                        borderColor: "rgba(255, 215, 0, 1)",
                        data: null,
                        fill: false,
                    },
                ],
            },
            options: {
                responsive: true,
                maintainAspectRatio: true,
                aspectRatio: 5,
                scales: {
                    y: {
                        beginAtZero: true,
                        max: 10,
                    },
                },
            },
        });
    });

    onDestroy(() => clearInterval(timer));
</script>

<canvas bind:this={ctx} />
