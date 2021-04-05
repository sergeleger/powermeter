<script lang="ts">
    import { onMount } from "svelte";
    import type { PowerSummary } from "./powersummary";

    export let year: number;
    export let month: number = 1;

    let entries: PowerSummary[] = [];
    let mounted: boolean;
    onMount(() => (mounted = true));

    // update URL only when year and month are valid.
    let url: string;
    $: {
        if (year > 2000 && year < 2100 && month >= 1 && month <= 12) {
            // @ts-ignore: SERVICE_URL is included in rollup build.
            url = SERVICE_URL + `/${year}/${month}`;
        }
    }

    // fetch data.
    $: {
        const fetchPower = async () => {
            const resp = await fetch(url);
            entries = await resp.json();
        };
        if (mounted) {
            fetchPower();
        }
    }

    // calculate total power consumption.
    let total: number;
    $: total = entries
        ? entries.reduce((sum, current) => sum + current.consumption, 0)
        : 0;
</script>

<table>
    <thead>
        <tr><th>Date</th><th>Consumption</th></tr>
    </thead>
    <tbody>
        {#each entries as u}
            <tr>
                <td>{u.year}-{u.month}-{u.day}</td>
                <td>{u.consumption}</td>
            </tr>
        {/each}
        <tr>
            <td />
            <td>{total}</td>
        </tr>
    </tbody>
</table>
