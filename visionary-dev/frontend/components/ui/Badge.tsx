import Image from "next/image";

interface BadgeProps {
  label: string;
  imageSrc?: string;
  icon?: React.ReactNode;
  variant?: "outline" | "gray";
  className?: string;
}

export default function Badge({ label, imageSrc, icon, variant = "outline", className = "" }: BadgeProps) {
  const base = "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium";
  const variants = {
    outline: "border border-gray-200 text-gray-500 bg-white",
    gray: "bg-gray-100 text-gray-500",
  };
  return (
    <span className={`${base} ${variants[variant]} ${className}`}>
      {imageSrc && (
        <span className="relative w-4 h-4 shrink-0">
          <Image src={imageSrc} alt={label} fill className="object-contain" />
        </span>
      )}
      {icon && !imageSrc && icon}
      {label}
    </span>
  );
}
