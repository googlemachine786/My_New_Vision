"use client";

import dynamic from "next/dynamic";
import type { ApexOptions } from "apexcharts";

const ApexChart = dynamic(() => import("react-apexcharts"), { ssr: false });

export interface BarChartSeries {
  name: string;
  data: { x: string; y: number }[];
  color?: string;
}

interface BarChartProps {
  series: BarChartSeries[];
  height?: number;
  primaryColor?: string;
  secondaryColor?: string;
}

export default function BarChart({
  series,
  height = 240,
  primaryColor = "#2563EB",
  secondaryColor = "#93C5FD",
}: BarChartProps) {
  const options: ApexOptions = {
    colors: [primaryColor, secondaryColor],
    chart: {
      type: "bar",
      fontFamily: "Inter, sans-serif",
      toolbar: { show: false },
      background: "transparent",
    },
    plotOptions: {
      bar: {
        horizontal: false,
        columnWidth: "70%",
        borderRadiusApplication: "end",
        borderRadius: 6,
      },
    },
    dataLabels: { enabled: false },
    legend: { show: false },
    tooltip: {
      shared: true,
      intersect: false,
      style: { fontFamily: "Inter, sans-serif" },
    },
    stroke: { show: true, width: 0, colors: ["transparent"] },
    grid: {
      show: true,
      borderColor: "#F3F4F6",
      strokeDashArray: 4,
      padding: { left: 2, right: 2, top: -14 },
    },
    xaxis: {
      floating: false,
      labels: {
        show: true,
        style: { fontFamily: "Inter, sans-serif", fontSize: "12px", colors: "#9CA3AF" },
      },
      axisBorder: { show: false },
      axisTicks: { show: false },
    },
    yaxis: {
      labels: {
        style: { fontFamily: "Inter, sans-serif", fontSize: "12px", colors: "#9CA3AF" },
        formatter: (val) => `${val}%`,
      },
      min: 0,
      max: 100,
      tickAmount: 5,
    },
    fill: { opacity: 1 },
    states: {
      hover: { filter: { type: "darken" } },
    },
  };

  return (
    <ApexChart
      type="bar"
      series={series}
      options={options}
      height={height}
      width="100%"
    />
  );
}
