import { Injectable, Logger } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import * as nodemailer from "nodemailer";

@Injectable()
export class EmailService {
  private readonly logger = new Logger(EmailService.name);
  private transporter: nodemailer.Transporter;

  constructor(private config: ConfigService) {
    this.transporter = nodemailer.createTransport({
      host: this.config.get("SMTP_HOST"),
      port: Number(this.config.get("SMTP_PORT") ?? 587),
      secure: this.config.get("SMTP_SECURE") === "true",
      auth: {
        user: this.config.get("SMTP_USER"),
        pass: this.config.get("SMTP_PASS"),
      },
    });
  }

  async sendOtp(email: string, otp: string) {
    const from = this.config.get("SMTP_FROM") || "Visionary <noreply@visionary.com>";
    const frontendUrl = this.config.get("FRONTEND_URL") || "http://localhost:3000";

    await this.transporter.sendMail({
      from,
      to: email,
      subject: "Your Visionary verification code",
      html: `
        <div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
          <img src="${frontendUrl}/logo.png" alt="Visionary" width="48" style="margin-bottom: 24px;" />
          <h2 style="font-size: 24px; font-weight: 600; color: #111827; margin: 0 0 8px;">
            Verify your email
          </h2>
          <p style="color: #6B7280; margin: 0 0 32px;">
            Enter the code below to verify your Visionary account. It expires in 10 minutes.
          </p>
          <div style="letter-spacing: 12px; font-size: 36px; font-weight: 700; color: #2563EB;
                      background: #EFF6FF; border-radius: 12px; padding: 20px 24px;
                      text-align: center; margin-bottom: 32px;">
            ${otp}
          </div>
          <p style="color: #9CA3AF; font-size: 13px;">
            If you didn't request this, you can safely ignore this email.
          </p>
        </div>
      `,
    });

    this.logger.log(`OTP sent to ${email}`);
  }

  async sendDeletionConfirmation(email: string, name: string) {
    const from = this.config.get("SMTP_FROM") || "Visionary <noreply@visionary.com>";
    const frontendUrl = this.config.get("FRONTEND_URL") || "http://localhost:3000";

    await this.transporter.sendMail({
      from,
      to: email,
      subject: "Your Visionary account deletion request",
      html: `
        <div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 32px;">
          <img src="${frontendUrl}/logo.png" alt="Visionary" width="48" style="margin-bottom: 24px;" />
          <h2 style="font-size: 24px; font-weight: 600; color: #111827; margin: 0 0 8px;">
            Account Deletion Request Received
          </h2>
          <p style="color: #6B7280; margin: 0 0 16px;">
            Hi ${name},
          </p>
          <p style="color: #6B7280; margin: 0 0 16px;">
            We've received your request to delete your Visionary account.
          </p>
          <div style="background: #FEF2F2; border-radius: 12px; padding: 20px 24px; margin-bottom: 24px;">
            <p style="color: #DC2626; font-weight: 600; margin: 0 0 8px;">⚠ Important</p>
            <p style="color: #7F1D1D; margin: 0; font-size: 14px;">
              All your personal data, course progress, and certificates will be
              <strong>permanently deleted in 7 days</strong>. This action cannot be undone.
            </p>
          </div>
          <p style="color: #6B7280; font-size: 13px;">
            If you did not request this or changed your mind, please contact our support team immediately.
          </p>
        </div>
      `,
    });

    this.logger.log(`Deletion confirmation sent to ${email}`);
  }
}
