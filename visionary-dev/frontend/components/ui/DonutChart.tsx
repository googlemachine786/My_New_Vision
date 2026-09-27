"use client";

import dynamic from "next/dynamic";
import type { ApexOptions } from "apexcharts";

const ApexChart = dynamic(() => import("react-apexcharts"), { ssr: false });

interface DonutChartProps {
  series?: number[];
  labels?: string[];
  height?: number;
  centerLabel?: string;
  colors?: string[];
}

export default function DonutChart({
  series = [35.1, 23.5, 2.4, 5.4],
  labels = ["Direct", "Sponsor", "Affiliate", "Email marketing"],
  height = 280,
  centerLabel = "Overall Progress",
  colors = ["#2563EB", "#93C5FD", "#1D4ED8"],
}: DonutChartProps) {
  const options: ApexOptions = {
    series,
    colors,
    chart: {
      height,
      width: "100%",
      type: "donut",
      fontFamily: "Inter, sans-serif",
      toolbar: { show: false },
      background: "transparent",
    },
    stroke: {
      colors: ["transparent"],
    },
    plotOptions: {
      pie: {
        donut: {
          size: "80%",
          labels: {
            show: true,
            name: {
              show: true,
              fontFamily: "Inter, sans-serif",
              offsetY: 20,
            },
            total: {
              showAlways: true,
              show: true,
              label: centerLabel,
              fontFamily: "Inter, sans-serif",
              formatter: (w) => {
                const sum = w.globals.seriesTotals.reduce(
                  (a: number, b: number) => a + b,
                  0
                );
                return sum + "%";
              },
            },
            value: {
              show: true,
              fontFamily: "Inter, sans-serif",
              offsetY: -20,
              formatter: (value) => value + "%",
            },
          },
        },
      },
    },
    grid: { padding: { top: -2 } },
    labels,
    dataLabels: { enabled: false },
    legend: {
      show: false,
    },
    yaxis: {
      labels: { formatter: (value) => value + "%" },
    },
  };

  return <ApexChart type="donut" series={series} options={options} height={height} width="100%" />;
}
