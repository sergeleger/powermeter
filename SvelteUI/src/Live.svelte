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
		type ChartDataset,
	} from "chart.js";
	import { onDestroy, onMount } from "svelte";
	import { getDayConsumption, type PowerDetail } from "./powersummary";
	import { consumptionPlugin, type DataPoint } from "./consumptionPlugin";

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

	const kwhDataset: ChartDataset<"bar", DataPoint[]> = {
		type: "bar",
		label: "kWh",
		backgroundColor: "rgba(0, 128, 0, 0.6)",
		data: [],
		borderWidth: 1,
		parsing: {
			yAxisKey: "value",
		},
	};

	const movingAvgDataset: ChartDataset<"line", DataPoint[]> = {
		type: "line",
		label: "3-hour moving average",
		borderColor: "rgba(255, 215, 0, 1)",
		data: [],
		fill: false,
		parsing: {
			yAxisKey: "value",
		},
	};

	// refreshRateInMS sets the refresh rate of the live view in milliseconds
	export let refreshRateInMS: number = 60_000;

	let ctx: HTMLCanvasElement;
	let timer: ReturnType<typeof setTimeout>;
	let chart: Chart<"bar" | "line", DataPoint[], unknown>;

	async function fetchPower() {
		// fetch Power entries
		//const today: Date = new Date();
		const today: Date = new Date(2024, 6, 1);

		// fetch 3 days worth of data to ensure we are able to get 48 entries
		const queries: Promise<PowerDetail[]>[] = [];
		for (let i = 0; i < 3; i++) {
			queries.push(getDayConsumption(today, true));
			today.setDate(today.getDate() - 1);
		}

		const usage: Array<PowerDetail[]> = await Promise.all(queries);
		const allUsage = [...usage[2], ...usage[1], ...usage[0]];

		const entries: DataPoint[] = new Array(48);
		const labels: string[] = new Array(48);
		const now: Date = new Date();
		let j = allUsage.length - 1;
		for (let i = 47; i >= 0; i--) {
			const h = now.getHours();
			labels[i] = `${h}h`;

			entries[i] = { value: 0, details: [], x: i };
			for (; j >= 0 && allUsage[j].hour === h; j--) {
				entries[i].value += allUsage[j].consumption;
				entries[i].details!.push(allUsage[j]);
			}

			now.setHours(h - 1);
		}

		const avg = new Array<number>(3);
		chart.data.labels = labels;
		chart.data.datasets[0].data = entries;
		chart.data.datasets[1].data = entries.map((d, i) => {
			avg[i % 3] = d.value;
			const sum = avg.reduce((tot, v) => tot + v, 0);
			return { value: sum / 3, details: null, x: i };
		});
		chart.options.scales!.y!.max = Math.max(7, ...entries.map((e) => e.value));
		chart.update();
	}

	onMount(() => {
		fetchPower();
		timer = setInterval(() => fetchPower(), refreshRateInMS);

		chart = new Chart<"bar" | "line", DataPoint[], unknown>(ctx, {
			type: "bar",
			plugins: [consumptionPlugin],
			data: {
				labels: [],
				datasets: [kwhDataset, movingAvgDataset],
			},
			options: {
				responsive: true,
				maintainAspectRatio: true,
				aspectRatio: 3,
				scales: {
					y: {
						beginAtZero: true,
						ticks: {
							stepSize: 1,
						},
					},
				},
			},
		});
	});

	onDestroy(() => clearInterval(timer));
</script>

<canvas bind:this={ctx} />
