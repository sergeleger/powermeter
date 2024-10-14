export interface PowerSummary {
	mete: number;
	year: number;
	consumption: number;
	month?: number;
	day?: number;
	hour?: number;
}

export interface PowerDetail {
	meter: number;
	year: number;
	month: number;
	day: number;
	seconds: number;
	consumption: number;
	hour: number;
}

export function getDayConsumption(date: Date, details = false): Promise<PowerDetail[]> {
	const y: number = date.getFullYear();
	const m: number = date.getMonth() + 1;
	const d: number = date.getDate();
	const url = `/api/${y}/${m}/${d}?details=${details}`;

	return new Promise<PowerDetail[]>((accept, reject) => {
		fetch(url)
			.then((resp) => resp.json())
			.then((data) => {
				data?.forEach((d: PowerDetail) => (d.hour = Math.floor(d.seconds / 3600)));
				accept(data ?? []);
			})
			.catch(reject);
	});
}
