import { request } from './request';
import { API_BASE_URL } from './config';
import { tokenManager } from './token';

export interface Attachment {
  id: number;
  project_id: number;
  conversation_id: number | null;
  file_name: string;
  original_name: string;
  file_path: string;
  file_size: number;
  content_type: string;
  type: 'image' | 'pdf';
  sort_order: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface UploadAttachmentRequest {
  file: File;
}

export interface AttachmentApiResponse<T = any> {
  message: string;
  data: T;
}

export const attachmentApi = {
  // Upload attachment (no conversation association yet)
  async uploadAttachment(file: File, projectId: number): Promise<Attachment> {
    const formData = new FormData();
    formData.append('file', file);

    const token = tokenManager.getToken();
    const response = await fetch(`${API_BASE_URL}/projects/${projectId}/attachments/upload`, {
      method: 'POST',
      body: formData,
      headers: {
        ...(token && { 'Authorization': `Bearer ${token}` }),
      },
    });

    if (!response.ok) {
      throw new Error(`Upload failed: ${response.statusText}`);
    }

    const result: AttachmentApiResponse<Attachment> = await response.json();
    return result.data;
  },

  // Get attachments for a project (optionally filtered by conversation)
  async getProjectAttachments(projectId: number, conversationId?: number): Promise<Attachment[]> {
    const queryParam = conversationId ? `?conversation_id=${conversationId}` : '';
    const response = await request<AttachmentApiResponse<Attachment[]>>(
      `/projects/${projectId}/attachments${queryParam}`
    );
    return response.data;
  },

  // Get attachments for a conversation (legacy method for backward compatibility)
  async getConversationAttachments(conversationId: number, projectId: number): Promise<Attachment[]> {
    return this.getProjectAttachments(projectId, conversationId);
  },

  // Get preview with authentication for images
  async getPreviewBlob(id: number, projectId: number): Promise<string> {
    const token = tokenManager.getToken();
    const response = await fetch(`${API_BASE_URL}/projects/${projectId}/attachments/${id}/preview`, {
      headers: {
        ...(token && { 'Authorization': `Bearer ${token}` }),
      },
    });
    
    if (!response.ok) {
      throw new Error(`Preview failed: ${response.statusText}`);
    }
    
    const blob = await response.blob();
    return window.URL.createObjectURL(blob);
  },

  // Delete attachment
  async deleteAttachment(id: number, projectId: number): Promise<void> {
    await request(`/projects/${projectId}/attachments/${id}`, {
      method: 'DELETE',
    });
  },

  // Download attachment file
  async downloadAttachment(id: number, filename: string, projectId: number): Promise<void> {
    const token = tokenManager.getToken();
    const response = await fetch(`${API_BASE_URL}/projects/${projectId}/attachments/${id}/download`, {
      headers: {
        ...(token && { 'Authorization': `Bearer ${token}` }),
      },
    });
    
    if (!response.ok) {
      throw new Error(`Download failed: ${response.statusText}`);
    }
    
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
};
