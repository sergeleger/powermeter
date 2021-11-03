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
		Legend,
	} from "chart.js";
	import type { ChartDataset } from "chart.js";
	import { onMount } from "svelte";
	import type { PowerSummary } from "./powersummary";
	import PowerMeter from "./PowerMeter.svelte";

	Chart.register(
		LinearScale,
		BarController,
		LineController,
		CategoryScale,
		BarElement,
		Title,
		Tooltip,
		PointElement,
		LineElement,
		Legend
	);

	let ctx: HTMLCanvasElement;
	let chart: Chart;

	const colors = [
		"rgba(255, 99, 132, 0.5)",
		"rgba(54, 162, 235, 0.5)",
		"rgba(255, 206, 86, 0.5)",
		"rgba(75, 192, 192, 0.5)",
		"rgba(153, 102, 255, 0.5)",
		"rgba(255, 159, 64, 0.5)",
	];

	onMount(async () => {
		// @ts-ignore: SERVICE_URL is included in rollup build.
		const serviceURL = SERVICE_URL;

		chart = new Chart(ctx, {
			type: "bar",
			data: {
				datasets: [],
			},
			options: {
				responsive: true,
				maintainAspectRatio: true,
				aspectRatio: 4,
				scales: {
					y: {
						beginAtZero: true,
					},
				},
			},
		});

		const years: PowerSummary[] = await fetch(serviceURL).then(async (r) => r.json());

		var yearDetail = await Promise.all(
			years.map((y) => fetch(serviceURL + `/${y.year}`).then((response) => response.json()))
		);

		yearDetail.map((usage, i) => {
			const dataset: ChartDataset<"bar", number[]> = {
				data: new Array<number>(12),
				backgroundColor: colors[i % colors.length],
				label: `${usage[0].year}`,
			};

			dataset.data = usage.map((u) => (dataset.data[u.month - 1] = u.consumption));

			chart.data.datasets.push(dataset);
		});

		chart.data.labels = [
			"January",
			"February",
			"March",
			"April",
			"May",
			"June",
			"July",
			"August",
			"September",
			"October",
			"November",
			"December",
		];

		chart.update();
	});
</script>

<canvas bind:this={ctx} />
