<script lang="ts">
	import { onDestroy, onMount } from "svelte";
	import type { PowerDetail } from "./powersummary";

	let refreshRateInMS: number = 60_000;

	let timer: ReturnType<typeof setTimeout>;
	let speed: number = 0;

	onMount(async () => {
		fetchSpeed();
		timer = setInterval(fetchSpeed, refreshRateInMS);
	});
	onDestroy(() => clearInterval(timer));

	async function fetchSpeed(): Promise<void> {
		const today: Date = new Date();
		const y: number = today.getFullYear();
		const m: number = today.getMonth() + 1;
		const d: number = today.getDate();

		const url = `/api/${y}/${m}/${d}?details=true`;
		const resp = await fetch(url);
		const usage: PowerDetail[] = await resp.json();

		const n = usage?.length ?? 0;
		let start = 0;
		let end = 0;
		switch (true) {
			case n === 0:
				speed = 0;
				return;

			case n === 1:
				end = usage[0].seconds;
				break;

			case n > 1:
				start = usage[n - 2].seconds;
				end = usage[n - 1].seconds;
				break;
		}

		let delta = end - start;

		// if we are currently past the previous time span, use the current time to calculate the
		// actual speed
		const now = today.getMinutes() * 60 + today.getHours() * 3600 + today.getSeconds();
		if (end + delta < now) {
			delta = now - start;
		}

		speed = 3600 / (delta / usage[n - 1].consumption);
	}
</script>

<h2>Speed: {speed.toFixed(2)} kWh / hour</h2>
