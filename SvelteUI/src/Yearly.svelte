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
		type Color,
	} from "chart.js";
	import type { ChartDataset } from "chart.js";
	import { onMount } from "svelte";
	import type { PowerSummary } from "./powersummary";
	import { interpolateSinebow as colorFn } from "d3-scale-chromatic";

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

	onMount(async () => {
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

		const years: PowerSummary[] = await fetch("/api").then(async (r) => r.json());

		var yearDetail = await Promise.all(
			years.map((y) => fetch(`/api/${y.year}`).then((response) => response.json()))
		);

		yearDetail.map((usage, i) => {
			const color: Color = colorWithAlpha(parseRGB(colorFn(i / yearDetail.length)), 0.75);

			const dataset: ChartDataset<"bar", number[]> = {
				data: new Array<number>(12),
				backgroundColor: color,
				label: `${usage[0].year}`,
			};

			dataset.data = usage.map(
				(u: PowerSummary) => (dataset.data[u.month! - 1] = u.consumption)
			);

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

	// Converts RGB string to R,G,B components.
	function parseRGB(color: string): number[] {
		return color
			.replace(/[^\d,]/g, "")
			.split(",")
			.map((c) => parseInt(c));
	}

	// Adds transparency to the RGB color.
	function colorWithAlpha(rgb: number[], alpha: number): Color {
		return `rgba(${rgb[0]}, ${rgb[1]}, ${rgb[2]}, ${alpha})`;
	}
</script>

<canvas bind:this={ctx} />
