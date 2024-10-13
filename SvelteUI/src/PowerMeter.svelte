<script lang="ts">
	import { onMount } from "svelte";
	import type { PowerSummary } from "./powersummary";

	export let year: number;
	export let month: number = 1;

	const formatter = new Intl.NumberFormat("en-CA", {
		minimumIntegerDigits: 2,
	});

	let entries: PowerSummary[] = [];
	let mounted: boolean;
	onMount(() => (mounted = true));

	// update URL only when year and month are valid.
	let url: string;
	$: {
		if (year > 2000 && year < 2100 && month >= 1 && month <= 12) {
			url = `/api/${year}/${month}`;
		}
	}

	// fetch data.
	$: {
		const fetchPower = async () => {
			const resp = await fetch(url);
			const data = await resp.json();
			entries = data ?? [];
		};
		if (mounted) {
			fetchPower();
		}
	}

	// calculate total power consumption.
	let total: number;
	$: total = entries ? entries.reduce((sum, current) => sum + current.consumption, 0) : 0;
</script>

<table>
	<thead>
		<tr><th>Date</th><th>Consumption</th></tr>
	</thead>
	<tbody>
		{#each entries as u}
			<tr>
				<td>{u.year}-{formatter.format(u.month ?? 0)}-{formatter.format(u.day ?? 0)}</td>
				<td>{u.consumption}</td>
			</tr>
		{/each}
		<tr>
			<th scope="row">Total</th>
			<td>{total}</td>
		</tr>
	</tbody>
</table>
