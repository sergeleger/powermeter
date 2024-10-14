import type { Plugin, BarProps, BarOptions, ScriptableContext, Chart } from "chart.js";
import { type PowerDetail } from "./powersummary";

type BarElement = {
	$context: ScriptableContext<"bar">;
	options: BarOptions;
} & BarProps;

export interface DataPoint {
	value: number;
	details: PowerDetail[] | null;
	x: number;
}

export const consumptionPlugin: Plugin<"bar", {}> = {
	id: "consumption_plugin",

	afterDatasetDraw: function (chart, args, options) {
		if (args.index != 0) {
			return;
		}

		args.meta.data.forEach((unk: any) => {
			const b: BarElement = unk;
			const raw = b.$context.raw as DataPoint;
			if (isNaN(b.height) || raw.value === 0) {
				return;
			}

			drawTimeScale(chart, b, raw);
			drawUsage(chart, b, raw);
		});
	},
};

// drawTimeScale renders time-ticks every 10 minutes along the vertical bar element.
function drawTimeScale(chart: Chart<"bar">, bar: BarElement, dp: DataPoint) {
	const yAxis = chart.scales["y"];
	const width = getBarWidth(bar);
	const x = bar.x - width / 2;

	chart.ctx.save();
	chart.ctx.lineWidth = 2;
	chart.ctx.strokeStyle = "rgba(0, 128, 0, 1)"; //"rgba(101, 101, 101, 0.75)";
	chart.ctx.shadowOffsetX = 1;
	chart.ctx.shadowOffsetY = 1;
	chart.ctx.shadowColor = "rgba(255, 255, 255, 0.75)";

	for (let i = 15; i < 60; i += 15) {
		const y = yAxis.getPixelForValue((i / 60) * dp.value);

		chart.ctx.beginPath();
		chart.ctx.moveTo(x, y);
		chart.ctx.lineTo(x + width / 2, y);
		chart.ctx.stroke();
	}
	chart.ctx.restore();
}

function drawUsage(chart: Chart<"bar">, bar: BarElement, dp: DataPoint) {
	const yAxis = chart.scales["y"];

	const width = getBarWidth(bar);
	const x = bar.x;

	chart.ctx.save();
	chart.ctx.lineWidth = 1;
	chart.ctx.strokeStyle = "rgba(0, 128, 0, 1)";
	chart.ctx.fillStyle = "rgba(0, 128, 0, 1)";

	dp.details!.sort((a, b) => a.seconds - b.seconds)
		.map((p) => yAxis.getPixelForValue((p.seconds / 3600 - p.hour) * dp.value))
		.forEach((h) => {
			const region = new Path2D();
			region.moveTo(x + 1, h);
			region.lineTo(x + width / 2 - 1, h - 4);
			region.lineTo(x + width / 2 - 1, h + 4);
			region.lineTo(x + 1, h);
			region.closePath();
			chart.ctx.fill(region);
		});
	chart.ctx.restore();
}

function getBarWidth(bar: BarElement): number {
	if (typeof bar.options.borderWidth === "number") {
		return bar.width - bar.options.borderWidth * 2;
	}

	return bar.width - bar.options.borderWidth.left! - bar.options.borderWidth.right!;
}
