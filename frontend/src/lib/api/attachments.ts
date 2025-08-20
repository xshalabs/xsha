import { apiClient } from './request';

export interface Attachment {
  id: number;
  conversation_id: number;
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
  conversation_id: number;
  file: File;
}

export interface AttachmentApiResponse<T = any> {
  message: string;
  data: T;
}

export const attachmentApi = {
  // Upload attachment
  async uploadAttachment(conversationId: number, file: File): Promise<Attachment> {
    const formData = new FormData();
    formData.append('conversation_id', conversationId.toString());
    formData.append('file', file);

    const response = await apiClient.post<AttachmentApiResponse<Attachment>>(
      '/attachments/upload',
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
    return response.data.data;
  },

  // Get attachment info
  async getAttachment(id: number): Promise<Attachment> {
    const response = await apiClient.get<AttachmentApiResponse<Attachment>>(
      `/attachments/${id}`
    );
    return response.data.data;
  },

  // Get attachments for a conversation
  async getConversationAttachments(conversationId: number): Promise<Attachment[]> {
    const response = await apiClient.get<AttachmentApiResponse<Attachment[]>>(
      `/attachments?conversation_id=${conversationId}`
    );
    return response.data.data;
  },

  // Download attachment
  getDownloadUrl(id: number): string {
    return `${apiClient.defaults.baseURL}/attachments/${id}/download`;
  },

  // Preview attachment (for images)
  getPreviewUrl(id: number): string {
    return `${apiClient.defaults.baseURL}/attachments/${id}/preview`;
  },

  // Delete attachment
  async deleteAttachment(id: number): Promise<void> {
    await apiClient.delete(`/attachments/${id}`);
  },

  // Download attachment file
  async downloadAttachment(id: number, filename: string): Promise<void> {
    const response = await apiClient.get(`/attachments/${id}/download`, {
      responseType: 'blob',
    });
    
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
};
