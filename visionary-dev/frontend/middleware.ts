import { NextRequest, NextResponse } from "next/server";

function decodeJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const base64 = token.split(".")[1];
    const json = Buffer.from(base64, "base64").toString("utf-8");
    return JSON.parse(json);
  } catch {
    return null;
  }
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("auth_token")?.value;

  const isOtpPage = pathname.startsWith("/signup/otp");
  const isAuthPage =
    (pathname.startsWith("/login") || pathname.startsWith("/signup")) && !isOtpPage;
  const isOnboardingPage = pathname.startsWith("/onboarding");
  const isCallbackPage = pathname.startsWith("/auth/callback");
  const isProtectedPage =
    pathname.startsWith("/dashboard") || pathname.startsWith("/chat");

  // Always allow callback and OTP page
  if (isCallbackPage || isOtpPage) return NextResponse.next();

  if (!token) {
    if (isAuthPage) return NextResponse.next();
    return NextResponse.redirect(new URL("/login", request.url));
  }

  const payload = decodeJwtPayload(token);
  if (!payload) {
    const res = NextResponse.redirect(new URL("/login", request.url));
    res.cookies.delete("auth_token");
    return res;
  }

  const onboardingCompleted = payload.onboardingCompleted === true;

  // Logged in + onboarding done → block auth pages
  if (isAuthPage && onboardingCompleted) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  // Logged in + onboarding not done → force onboarding
  if (!onboardingCompleted && (isProtectedPage || isAuthPage)) {
    return NextResponse.redirect(new URL("/onboarding", request.url));
  }

  // Logged in + onboarding done → block onboarding page
  if (onboardingCompleted && isOnboardingPage) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/login",
    "/login/:path*",
    "/signup",
    "/signup/:path*",
    "/onboarding",
    "/onboarding/:path*",
    "/dashboard",
    "/dashboard/:path*",
    "/chat",
    "/chat/:path*",
  ],
};
