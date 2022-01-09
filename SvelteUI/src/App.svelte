<script lang="ts">
	import PowerMeter from "./PowerMeter.svelte";
	import Live from "./Live.svelte";
	import Yearly from "./Yearly.svelte";
	import Speed from "./Speed.svelte";

	let year: number = new Date().getFullYear();
	let month: number = new Date().getMonth() + 1;

	function next() {
		month++;
		if (month <= 12) {
			return;
		}

		year++;
		month = 1;
	}

	function prev() {
		month--;
		if (month >= 1) {
			return;
		}

		year--;
		month = 12;
	}
</script>

<main>
	<h1>PowerMeter</h1>

	<Speed />

	<Live />

	<Yearly />

	<button on:click={prev}>&lt;</button>
	<label>
		<input type="number" bind:value={year} />
	</label>

	<label>
		<input type="number" bind:value={month} />
	</label>
	<button on:click={next}>&gt;</button>

	<PowerMeter {year} {month} />
</main>

<style>
	main {
		text-align: center;
		padding: 1em;
		max-width: 240px;
		margin: 0 auto;
	}

	h1 {
		color: #ff3e00;
		/* text-transform: uppercase; */
		font-size: 4em;
		font-weight: 100;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}

	label,
	button {
		display: inline;
		font-size: 1.5em;
	}

	input {
		width: 4em;
		text-align: center;
	}
</style>
