import Image from "next/image";

interface StatCardProps {
  imagePath: string;
  label: string;
  number: number;
}

export default function StatCard({ imagePath, label, number }: StatCardProps) {
  return (
    <div className="bg-white border border-gray-100 rounded-2xl px-6 py-4 flex items-center gap-4 flex-1">
      <div className="relative w-12 h-12 shrink-0">
        <Image src={imagePath} alt={label} fill className="object-contain" />
      </div>
      <span className="text-sm text-gray-500 flex-1">{label}</span>
      <span className="text-2xl font-bold text-gray-900">{number}</span>
    </div>
  );
}
